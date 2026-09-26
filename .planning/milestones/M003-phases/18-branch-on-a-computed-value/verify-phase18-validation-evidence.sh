#!/bin/sh
set -eu

phase_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(git -C "$phase_dir/../../.." rev-parse --show-toplevel)
python3 - "$phase_dir/18-VALIDATION.md" "$repo_root/.github/workflows/ci.yml" "${1:---measurements}" <<'PY'
import math
import pathlib
import re
import subprocess
import sys

record_path, ci_path, mode = map(pathlib.Path, sys.argv[1:])
mode = str(mode)
if mode not in ("--measurements", "--final"):
    raise SystemExit("usage: verify-phase18-validation-evidence.sh [--measurements|--final]")
text = record_path.read_text()
blocks = re.findall(r"```text\s*\n(.*?)\n```", text, re.S)
if not blocks:
    raise SystemExit("missing fenced Phase 18 evidence record")
values = {}
for line in blocks[-1].splitlines():
    if not line.strip():
        continue
    if "=" not in line:
        raise SystemExit(f"invalid evidence line: {line}")
    key, value = line.split("=", 1)
    values[key.strip()] = value.strip()

def require(key):
    value = values.get(key, "")
    if not value:
        raise SystemExit(f"missing {key}")
    return value

if require("phase18_evidence_version") != "1":
    raise SystemExit("unsupported evidence version")
for key in ("host", "go_version", "clang_version"):
    require(key)
commands = {
    "focused": "go test ./internal/compiler/session -count=1",
    "vet": "go vet ./...",
    "build": "go build ./...",
    "full_test": "go test ./... -count=1",
    "race": "go test -race ./... -count=1",
}
for lane, command in commands.items():
    if require(f"{lane}_command") != command:
        raise SystemExit(f"{lane}_command must equal {command!r}")
    for kind in ("cold", "warm"):
        prefix = f"{lane}_{kind}"
        raw = require(f"{prefix}_seconds")
        samples = [float(item) for item in raw.split(",")]
        count = int(require(f"{prefix}_count"))
        if len(samples) < 3 or count != len(samples):
            raise SystemExit(f"{prefix}: need >=3 samples and matching count")
        if any(not math.isfinite(item) or item <= 0 for item in samples):
            raise SystemExit(f"{prefix}: samples must be positive finite seconds")
        ordered = sorted(samples)
        expected = {
            "min": ordered[0],
            "median": ordered[len(ordered)//2] if len(ordered) % 2 else sum(ordered[len(ordered)//2-1:len(ordered)//2+1])/2,
            "max": ordered[-1],
        }
        for stat, number in expected.items():
            actual = float(require(f"{prefix}_{stat}_seconds"))
            if not math.isclose(actual, number, rel_tol=0, abs_tol=0.001):
                raise SystemExit(f"{prefix}_{stat}_seconds={actual} does not match {number}")

before = require("ci_before_blob")
current = subprocess.check_output(["git", "hash-object", str(ci_path)], text=True).strip()
if before != current:
    raise SystemExit(f"CI blob changed: recorded {before}, current {current}")
if mode == "--final":
    disposition = require("ci_disposition")
    rationale = require("ci_decision_rationale")
    command = require("ci_command")
    if disposition == "not_added":
        if command != "none":
            raise SystemExit("not_added disposition requires ci_command=none")
        if len(rationale) < 40:
            raise SystemExit("not_added rationale must give a specific evidence-based tradeoff")
    elif disposition == "added":
        if command not in commands.values():
            raise SystemExit("added ci_command must exactly match a measured command")
        workflow = ci_path.read_text()
        checks = re.search(r"(?ms)^  checks:.*?(?=^  [A-Za-z0-9_-]+:|\Z)", workflow)
        if not checks or command not in checks.group(0):
            raise SystemExit("exact ci_command is absent from the existing checks job")
        if "ubuntu-latest" not in checks.group(0) or "macos-latest" not in checks.group(0):
            raise SystemExit("checks job must retain Ubuntu and macOS matrix hosts")
    else:
        raise SystemExit("ci_disposition must be added or not_added")
print(f"Phase 18 {mode[2:]} evidence verified")
PY
