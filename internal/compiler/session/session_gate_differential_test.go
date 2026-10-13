package session_test

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

type phase33CommandResult struct {
	exit           int
	stdout, stderr string
}

var phase33SHA = regexp.MustCompile(`[0-9a-f]{40}`)

func phase33RunAt(t *testing.T, cwd, script string, args ...string) phase33CommandResult {
	return phase33RunAtEnv(t, cwd, script, nil, args...)
}

func phase33RunAtEnv(t *testing.T, cwd, script string, environment []string, args ...string) phase33CommandResult {
	t.Helper()
	full := append([]string{"-c", `cd "$0" && script=$1 && shift && exec sh "$script" "$@"`, cwd, script}, args...)
	result, err := testsupport.RunCLIErr(context.Background(), "sh", environment, full...)
	if err != nil && result.Exit == 0 {
		t.Fatalf("spawn %s: %v", script, err)
	}
	return phase33CommandResult{result.Exit, string(result.Stdout), string(result.Stderr)}
}

func TestFreezeGateDifferential(t *testing.T) {
	t.Setenv("GIT_AUTHOR_DATE", "2000-01-01T00:00:00Z")
	t.Setenv("GIT_COMMITTER_DATE", "2000-01-01T00:00:00Z")
	read := func(path string) []byte {
		data, err := os.ReadFile(testsupport.ProjectPath(filepath.FromSlash(path)))
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	dir := t.TempDir()
	write := func(name string, data []byte) {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("cmd/a.go", []byte("package a\n"))
	write("scripts/codedigest.sh", read("scripts/codedigest.sh"))
	write("scripts/check-freeze.sh", read("scripts/check-freeze.sh"))
	write("testdata/receipts/M006.manifest", read("testdata/receipts/M006.manifest"))
	phase32Git(t, dir, "init", "-q")
	phase32Git(t, dir, "add", "-A")
	phase32Git(t, dir, "commit", "-qm", "base")
	base := phase32GateHead(t, dir)
	write("cmd/a.go", []byte("package a // freeze\n"))
	phase32Git(t, dir, "commit", "-qam", "freeze")
	freeze := phase32GateHead(t, dir)
	write(".planning/x.md", []byte("plan\n"))
	phase32Git(t, dir, "add", "-A")
	phase32Git(t, dir, "commit", "-qm", "planning")
	planning := phase32GateHead(t, dir)
	write("cmd/README.md", []byte("docs\n"))
	phase32Git(t, dir, "add", "-A")
	phase32Git(t, dir, "commit", "-qm", "markdown")
	markdown := phase32GateHead(t, dir)
	write("cmd/a.go", []byte("package a // changed\n"))
	phase32Git(t, dir, "commit", "-qam", "code")
	changed := phase32GateHead(t, dir)
	phase32Git(t, dir, "checkout", "-q", "--detach", base)
	write("other.txt", []byte("sibling\n"))
	phase32Git(t, dir, "add", "-A")
	phase32Git(t, dir, "commit", "-qm", "sibling")
	sibling := phase32GateHead(t, dir)
	new := filepath.Join(dir, "scripts/check-freeze.sh")
	manifest := "testdata/receipts/M006.manifest"
	rows := []struct {
		name, reason     string
		oldArgs, newArgs []string
	}{
		{"usage", "reason=usage", nil, []string{manifest}},
		{"bad_sha", "reason=bad_sha", []string{freeze, "BAD"}, []string{manifest, freeze, "BAD"}},
		{"bad_same_as", "reason=bad_sha", []string{freeze, planning, "BAD"}, []string{manifest, freeze, planning, "BAD"}},
		{"sha_mismatch", "reason=sha_mismatch", []string{freeze, planning, freeze}, []string{manifest, freeze, planning, freeze}},
		{"unknown_commit", "reason=unknown_commit", []string{freeze, strings.Repeat("0", 40)}, []string{manifest, freeze, strings.Repeat("0", 40)}},
		{"not_descendant", "reason=not_descendant", []string{freeze, sibling}, []string{manifest, freeze, sibling}},
		{"digest_change", "reason=digest_path_changed_after_freeze", []string{freeze, changed}, []string{manifest, freeze, changed}},
		{"md_only_pass", "status=pass", []string{freeze, markdown, markdown}, []string{manifest, freeze, markdown, markdown}},
		{"planning_only_pass", "status=pass", []string{freeze, planning, planning}, []string{manifest, freeze, planning, planning}},
		{"same_commit_pass", "status=pass", []string{freeze, freeze}, []string{manifest, freeze, freeze}},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			got := phase33RunAt(t, dir, new, row.newArgs...)
			got.stdout = strings.ReplaceAll(got.stdout, dir, "@TMP@")
			got.stderr = strings.ReplaceAll(got.stderr, dir, "@TMP@")
			got.stdout = phase33SHA.ReplaceAllString(got.stdout, "@SHA@")
			got.stderr = phase33SHA.ReplaceAllString(got.stderr, "@SHA@")
			want := freezeGateGoldens[row.name]
			if got != want {
				t.Fatalf("golden mismatch\n got=%#v\nwant=%#v", got, want)
			}
		})
	}
}

func TestReceiptGateDifferential(t *testing.T) {
	root := testsupport.ProjectPath()
	newScript := "scripts/check-receipts.sh"
	manifest := testsupport.ProjectPath("testdata", "receipts", "M006.manifest")
	mac := phase32ReceiptLog(t, "macOS", phase32GateDigest)
	linux := phase32ReceiptLog(t, "Linux", phase32GateDigest)
	type row struct {
		name, want string
		a, b       string
		args       int
	}
	rows := []row{
		{"usage", "usage_two_logs_required", mac, "", 1},
		{"host_missing", "host_missing_", mac, "\x00missing", 2},
		{"log_empty", "log_empty_", mac, "", 2},
		{"failed_or_unavailable_row", "failed_or_unavailable_row_", mac, strings.Replace(linux, "status=pass", "status=fail", 1), 2},
		{"identity_row_count", "identity_row_count_0_", mac, dropReceiptRows(t, linux, "phase32 host identity "), 2},
		{"result_row_count", "result_row_count_2_", mac, linux + "phase32 host result host=Linux\n", 2},
		{"identity_not_first", "identity_not_first_", mac, "phase32 toolchain go=\"g\" clang=\"c\"\n" + linux, 2},
		{"result_not_last", "result_not_last_", mac, linux + "phase32 toolchain go=\"g\" clang=\"c\"\n", 2},
		{"unknown_phase32_row", "unknown_phase32_row_", mac, strings.Replace(linux, "phase32 host result ", "phase32 forged row\nphase32 host result ", 1), 2},
		{"toolchain_row_count", "toolchain_row_count_", mac, dropReceiptRows(t, linux, "phase32 toolchain "), 2},
		{"toolchain_row_malformed", "toolchain_row_malformed_", mac, strings.Replace(linux, "go=\"go version go1.24.0\"", "go=\"\"", 1), 2},
		{"unknown_host", "unknown_host_", mac, strings.Replace(strings.Replace(linux, "host=Linux", "host=Plan9", 1), "host=Linux", "host=Plan9", 1), 2},
		{"arch_missing", "arch_missing_", mac, strings.Replace(linux, " arch=x86_64", "", 1), 2},
		{"target_missing", "target_missing_", mac, strings.Replace(linux, " clang_target=x86_64-pc-linux-gnu", "", 1), 2},
		{"go_host_missing", "go_host_missing_", mac, strings.Replace(linux, " go=linux/amd64", "", 1), 2},
		{"headsha_missing", "headsha_missing_", mac, strings.Replace(linux, " headsha="+phase32GateHeadSha, "", 1), 2},
		{"digest_mismatch", "digest_mismatch_", strings.Replace(mac, "codedigest="+phase32GateDigest, "codedigest="+strings.Repeat("cd", 32), 1), strings.Replace(linux, "codedigest="+phase32GateDigest, "codedigest="+strings.Repeat("cd", 32), 1), 2},
		{"tree_not_clean", "tree_not_clean_", mac, strings.Replace(linux, "tree=clean", "tree=dirty", 1), 2},
		{"result_host_mismatch", "result_host_mismatch_", mac, strings.Replace(linux, "phase32 host result host=Linux", "phase32 host result host=macOS", 1), 2},
		{"result_digest_mismatch", "result_digest_mismatch_", mac, strings.Replace(linux, "phase32 host result host=Linux codedigest="+phase32GateDigest, "phase32 host result host=Linux codedigest="+strings.Repeat("cd", 32), 1), 2},
		{"result_tree_not_clean", "result_tree_not_clean_", mac, strings.Replace(linux, "phase32 host result host=Linux codedigest="+phase32GateDigest+" tree=clean", "phase32 host result host=Linux codedigest="+phase32GateDigest+" tree=dirty", 1), 2},
		{"result_not_pass", "result_not_pass_", mac, strings.Replace(linux, "programs="+strings.Join(phase32ReceiptPrograms, ",")+" status=pass", "programs="+strings.Join(phase32ReceiptPrograms, ",")+" status=skipped", 1), 2},
		{"result_programs_mismatch", "result_programs_mismatch_", mac, strings.Replace(linux, "programs="+strings.Join(phase32ReceiptPrograms, ","), "programs=other", 1), 2},
		{"program_missing", "program_missing_adler_low_", mac, dropReceiptRows(t, linux, "phase32 case program=adler_low "), 2},
		{"result_control_count", "result_control_count_", mac, strings.Replace(linux, "controls=4", "controls=3", 1), 2},
		{"result_case_count", "result_case_count_", mac, strings.Replace(linux, "cases=39", "cases=38", 1), 2},
		{"case_row_count", "case_row_count_", mac, strings.Replace(linux, "phase32 host result ", strings.Split(linux, "\n")[2]+"\nphase32 host result ", 1), 2},
		{"case_missing_or_altered", "case_missing_or_altered_checksum_empty_", mac, strings.Replace(linux, "stdout=exact", "stdout=wrong", 1), 2},
		{"control_row_count", "control_row_count_", mac, dropReceiptRows(t, linux, "phase32 control name=codedigest "), 2},
		{"control_missing_or_altered", "control_missing_or_altered_guide_", mac, strings.Replace(linux, "name=guide package=./cmd/schway", "name=guide package=", 1), 2},
		{"log_digest_differs", "log_digest_differs", mac, phase32ReceiptLog(t, "Linux", strings.Repeat("cd", 32)), 2},
		{"duplicate_host", "duplicate_host_macOS", mac, mac, 2},
	}
	for _, tc := range rows {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			paths := []string{filepath.Join(dir, "a.log"), filepath.Join(dir, "b.log")}
			if err := os.WriteFile(paths[0], []byte(tc.a), 0o600); err != nil {
				t.Fatal(err)
			}
			if tc.b != "\x00missing" {
				if err := os.WriteFile(paths[1], []byte(tc.b), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			env := envWith("EXPECTED_CODEDIGEST", phase32GateDigest)
			args := append([]string{manifest}, paths[:tc.args]...)
			got := phase33RunAtEnv(t, root, newScript, env, args...)
			got.stdout = strings.ReplaceAll(got.stdout, dir, "@TMP@")
			got.stderr = strings.ReplaceAll(got.stderr, dir, "@TMP@")
			got.stdout = phase33SHA.ReplaceAllString(got.stdout, "@SHA@")
			got.stderr = phase33SHA.ReplaceAllString(got.stderr, "@SHA@")
			if got != receiptGateGoldens[tc.name] {
				t.Fatalf("golden mismatch\n got=%#v\nwant=%#v", got, receiptGateGoldens[tc.name])
			}
		})
	}
	for _, name := range []string{"case_table_empty", "programs_line_missing", "program_without_table_entry"} {
		t.Run(name, func(t *testing.T) {
			temp := t.TempDir()
			mText := string(mustRead(t, manifest))
			switch name {
			case "case_table_empty":
				mText = stripCaseRows(mText)
			case "programs_line_missing":
				mText = strings.Replace(mText, "programs="+strings.Join(phase32ReceiptPrograms, ",")+"\n", "", 1)
			case "program_without_table_entry":
				mText = strings.Replace(mText, "programs="+strings.Join(phase32ReceiptPrograms, ","), "programs="+strings.Join(phase32ReceiptPrograms, ",")+",bogus", 1)
				mac = strings.Replace(mac, "programs="+strings.Join(phase32ReceiptPrograms, ","), "programs="+strings.Join(phase32ReceiptPrograms, ",")+",bogus", 1)
				linux = strings.Replace(linux, "programs="+strings.Join(phase32ReceiptPrograms, ","), "programs="+strings.Join(phase32ReceiptPrograms, ",")+",bogus", 1)
			}
			variant := filepath.Join(temp, "variant.manifest")
			if err := os.WriteFile(variant, []byte(mText), 0o600); err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			a, b := filepath.Join(dir, "a.log"), filepath.Join(dir, "b.log")
			for path, content := range map[string]string{a: mac, b: linux} {
				if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			got := phase33RunAtEnv(t, root, newScript, envWith("EXPECTED_CODEDIGEST", phase32GateDigest), variant, a, b)
			got.stdout = strings.ReplaceAll(got.stdout, dir, "@TMP@")
			got.stderr = strings.ReplaceAll(got.stderr, dir, "@TMP@")
			got.stdout = strings.ReplaceAll(got.stdout, temp, "@TMP@")
			got.stderr = strings.ReplaceAll(got.stderr, temp, "@TMP@")
			if got != receiptGateGoldens[name] {
				t.Fatalf("golden mismatch: got=%#v want=%#v", got, receiptGateGoldens[name])
			}
		})
	}

}

var freezeGateGoldens = map[string]phase33CommandResult{
	"usage":              {exit: 1, stdout: "", stderr: "phase32 freeze gate: status=fail reason=usage\n"},
	"bad_sha":            {exit: 1, stdout: "", stderr: "phase32 freeze gate: status=fail reason=bad_sha\n"},
	"bad_same_as":        {exit: 1, stdout: "", stderr: "phase32 freeze gate: status=fail reason=bad_sha\n"},
	"sha_mismatch":       {exit: 1, stdout: "", stderr: "phase32 freeze gate: status=fail reason=sha_mismatch\n"},
	"unknown_commit":     {exit: 1, stdout: "", stderr: "phase32 freeze gate: status=fail reason=unknown_commit\n"},
	"not_descendant":     {exit: 1, stdout: "", stderr: "phase32 freeze gate: status=fail reason=not_descendant\n"},
	"digest_change":      {exit: 1, stdout: "", stderr: "phase32 freeze gate: status=fail reason=digest_path_changed_after_freeze\n"},
	"md_only_pass":       {exit: 0, stdout: "phase32 freeze gate: status=pass freeze=@SHA@ candidate=@SHA@\n", stderr: ""},
	"planning_only_pass": {exit: 0, stdout: "phase32 freeze gate: status=pass freeze=@SHA@ candidate=@SHA@\n", stderr: ""},
	"same_commit_pass":   {exit: 0, stdout: "phase32 freeze gate: status=pass freeze=@SHA@ candidate=@SHA@\n", stderr: ""},
}
var receiptGateGoldens = map[string]phase33CommandResult{
	"usage":                       {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=usage_two_logs_required\n"},
	"host_missing":                {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=host_missing_@TMP@/b.log\n"},
	"log_empty":                   {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=log_empty_@TMP@/b.log\n"},
	"failed_or_unavailable_row":   {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=failed_or_unavailable_row_@TMP@/b.log\n"},
	"identity_row_count":          {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=identity_row_count_0_@TMP@/b.log\n"},
	"result_row_count":            {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=result_row_count_2_@TMP@/b.log\n"},
	"identity_not_first":          {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=identity_not_first_@TMP@/b.log\n"},
	"result_not_last":             {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=result_not_last_@TMP@/b.log\n"},
	"unknown_phase32_row":         {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=unknown_phase32_row_@TMP@/b.log\n"},
	"toolchain_row_count":         {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=toolchain_row_count_@TMP@/b.log\n"},
	"toolchain_row_malformed":     {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=toolchain_row_malformed_@TMP@/b.log\n"},
	"unknown_host":                {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=unknown_host_@TMP@/b.log\n"},
	"arch_missing":                {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=arch_missing_@TMP@/b.log\n"},
	"target_missing":              {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=target_missing_@TMP@/b.log\n"},
	"go_host_missing":             {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=go_host_missing_@TMP@/b.log\n"},
	"headsha_missing":             {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=headsha_missing_@TMP@/b.log\n"},
	"digest_mismatch":             {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=digest_mismatch_@TMP@/a.log\n"},
	"tree_not_clean":              {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=tree_not_clean_@TMP@/b.log\n"},
	"result_host_mismatch":        {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=result_host_mismatch_@TMP@/b.log\n"},
	"result_digest_mismatch":      {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=result_digest_mismatch_@TMP@/b.log\n"},
	"result_tree_not_clean":       {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=result_tree_not_clean_@TMP@/b.log\n"},
	"result_not_pass":             {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=result_not_pass_@TMP@/b.log\n"},
	"result_programs_mismatch":    {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=result_programs_mismatch_@TMP@/b.log\n"},
	"program_missing":             {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=program_missing_adler_low_@TMP@/b.log\n"},
	"result_control_count":        {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=result_control_count_@TMP@/b.log\n"},
	"result_case_count":           {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=result_case_count_@TMP@/b.log\n"},
	"case_row_count":              {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=case_row_count_@TMP@/b.log\n"},
	"case_missing_or_altered":     {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=case_missing_or_altered_checksum_empty_@TMP@/b.log\n"},
	"control_row_count":           {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=control_row_count_@TMP@/b.log\n"},
	"control_missing_or_altered":  {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=control_missing_or_altered_guide_@TMP@/b.log\n"},
	"log_digest_differs":          {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=log_digest_differs\n"},
	"duplicate_host":              {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=duplicate_host_macOS\n"},
	"case_table_empty":            {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=case_table_empty\n"},
	"programs_line_missing":       {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=programs_line_missing\n"},
	"program_without_table_entry": {exit: 1, stdout: "", stderr: "phase32 receipt gate: status=fail reason=program_without_table_entry_bogus\n"},
}
