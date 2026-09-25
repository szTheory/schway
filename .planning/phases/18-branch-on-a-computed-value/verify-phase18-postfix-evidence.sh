#!/bin/sh
set -eu

phase_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(git -C "$phase_dir/../../.." rev-parse --show-toplevel)
exec python3 - "$phase_dir" "$repo_root" "${1:---final}" <<'PY'
import hashlib
import os
import pathlib
import platform
import re
import subprocess
import sys
import time
from datetime import datetime, timezone

phase_dir = pathlib.Path(sys.argv[1])
repo_root = pathlib.Path(sys.argv[2])
mode = sys.argv[3]
if mode not in ("--record", "--final"):
    raise SystemExit("usage: verify-phase18-postfix-evidence.sh [--record|--final]")

validation_path = phase_dir / "18-VALIDATION.md"
log_path = phase_dir / "18-09-RUNS.log"
source_files = (
    "internal/compiler/cache/cache.go",
    "internal/compiler/cache/probe.go",
    "internal/compiler/cache/probe_test.go",
    "internal/compiler/native/native.go",
    "internal/compiler/native/native_test.go",
)
lanes = (
    ("default_1", "GOCACHE=/tmp/ai-lang-gocache go test ./... -count=1", ("go", "test", "./...", "-count=1"), True),
    ("default_2", "GOCACHE=/tmp/ai-lang-gocache go test ./... -count=1", ("go", "test", "./...", "-count=1"), True),
    ("default_3", "GOCACHE=/tmp/ai-lang-gocache go test ./... -count=1", ("go", "test", "./...", "-count=1"), True),
    ("parallel_4", "GOCACHE=/tmp/ai-lang-gocache go test -p=4 ./... -count=1", ("go", "test", "-p=4", "./...", "-count=1"), True),
    ("race", "GOCACHE=/tmp/ai-lang-gocache go test -race ./... -count=1", ("go", "test", "-race", "./...", "-count=1"), True),
    ("vet", "go vet ./...", ("go", "vet", "./..."), True),
    ("build", "go build ./...", ("go", "build", "./..."), True),
)


def output(command, *, env=None):
    return subprocess.check_output(command, cwd=repo_root, env=env, stderr=subprocess.STDOUT, text=True).strip()


def versions():
    go = output(("go", "version"))
    clang = output(("clang", "--version")).splitlines()[0]
    host = platform.platform()
    return {"host": host, "go_version": go, "clang_version": clang}


def git_blob(path):
    return output(("git", "rev-parse", f"HEAD:{path}"))


def parse_record(block):
    values = {}
    for line in block.splitlines():
        if not line.strip():
            continue
        if "=" not in line:
            raise SystemExit(f"invalid phase18-postfix line: {line}")
        key, value = line.split("=", 1)
        key = key.strip()
        if key in values:
            raise SystemExit(f"duplicate phase18-postfix key: {key}")
        values[key] = value.strip()
    return values


def latest_postfix(text):
    blocks = re.findall(r"```phase18-postfix\s*\n(.*?)\n```", text, re.S)
    if not blocks:
        raise SystemExit("missing fenced phase18-postfix record")
    return parse_record(blocks[-1])


def require(values, key):
    value = values.get(key, "")
    if not value:
        raise SystemExit(f"missing {key}")
    return value


def phase08_values(text):
    blocks = re.findall(r"```text\s*\n(.*?)\n```", text, re.S)
    if not blocks:
        raise SystemExit("missing Plan 08 text evidence record")
    values = {}
    for line in blocks[-1].splitlines():
        if "=" in line:
            key, value = line.split("=", 1)
            values[key.strip()] = value.strip()
    return values


if mode == "--record":
    metadata = versions()
    blobs = {path: git_blob(path) for path in source_files}
    attempt_id = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ") + f"-{os.getpid()}"
    log_path.parent.mkdir(parents=True, exist_ok=True)
    results = []
    with log_path.open("ab") as log:
        for index, (lane, display, argv, use_gocache) in enumerate(lanes, start=1):
            env = os.environ.copy()
            if use_gocache:
                env["GOCACHE"] = "/tmp/ai-lang-gocache"
            header = f"\n=== phase18-postfix attempt={attempt_id} run={index} lane={lane} command={display} ===\n".encode()
            offset = log.tell()
            log.write(header)
            log.flush()
            started = time.monotonic()
            try:
                completed = subprocess.run(argv, cwd=repo_root, env=env, stdout=subprocess.PIPE,
                                           stderr=subprocess.STDOUT, check=False)
                captured = completed.stdout
                exit_code = completed.returncode
            except OSError as exc:
                captured = f"could not launch command: {exc}\n".encode()
                exit_code = 127
            duration = time.monotonic() - started
            section = header + captured
            log.write(captured)
            log.flush()
            results.append({
                "lane": lane,
                "command": display,
                "exit": exit_code,
                "duration": duration,
                "offset": offset,
                "length": len(section),
                "digest": hashlib.sha256(section).hexdigest(),
            })
            print(f"{lane}: exit={exit_code} duration={duration:.3f}s bytes={len(section)}", flush=True)

    lines = ["version=1", f"attempt_id={attempt_id}", f"run_count={len(results)}"]
    lines.extend(f"{key}={value}" for key, value in metadata.items())
    for path, blob in blobs.items():
        lines.append(f"source_blob_{path}={blob}")
    for index, result in enumerate(results, start=1):
        prefix = f"run_{index}"
        for field in ("lane", "command", "exit", "duration", "offset", "length", "digest"):
            lines.append(f"{prefix}_{field}={result[field]}")
    record = "\n\n---\n\nPost-G-18-16 default-parallel and aggregate verification receipts.\n\n```phase18-postfix\n" + "\n".join(lines) + "\n```\n"
    with validation_path.open("ab") as validation:
        validation.write(record.encode())
    if any(result["exit"] != 0 for result in results):
        raise SystemExit("one or more post-fix evidence lanes failed; inspect 18-09-RUNS.log")
    print("Recorded seven post-fix evidence lanes")
else:
    validation = validation_path.read_text()
    values = latest_postfix(validation)
    if require(values, "version") != "1" or require(values, "run_count") != "7":
        raise SystemExit("phase18-postfix record must contain exactly seven lane slots")
    attempt_id = require(values, "attempt_id")
    current_versions = versions()
    for key, current in current_versions.items():
        if require(values, key) != current:
            raise SystemExit(f"stale {key}: recorded {values[key]!r}, current {current!r}")
    for path in source_files:
        recorded = require(values, f"source_blob_{path}")
        if recorded != git_blob(path):
            raise SystemExit(f"stale committed source blob for {path}")

    expected_lanes = [lane for lane, _, _, _ in lanes]
    expected_commands = [display for _, display, _, _ in lanes]
    log = log_path.read_bytes()
    sections = []
    default_ranges = []
    for index, (expected_lane, expected_command) in enumerate(zip(expected_lanes, expected_commands), start=1):
        prefix = f"run_{index}_"
        lane = require(values, prefix + "lane")
        command = require(values, prefix + "command")
        if lane != expected_lane or command != expected_command:
            raise SystemExit(f"run slot {index} must be {expected_lane} with its exact command")
        if int(require(values, prefix + "exit")) != 0:
            raise SystemExit(f"run slot {index} ({lane}) did not pass")
        duration = float(require(values, prefix + "duration"))
        if duration <= 0:
            raise SystemExit(f"run slot {index} has a nonpositive duration")
        offset = int(require(values, prefix + "offset"))
        length = int(require(values, prefix + "length"))
        if offset < 0 or length <= 0 or offset + length > len(log):
            raise SystemExit(f"run slot {index} output section is out of bounds")
        section = log[offset:offset + length]
        digest = require(values, prefix + "digest")
        if hashlib.sha256(section).hexdigest() != digest:
            raise SystemExit(f"run slot {index} output digest mismatch")
        marker = f"\n=== phase18-postfix attempt={attempt_id} run={index} lane={lane} command={command} ===\n".encode()
        if not section.startswith(marker):
            raise SystemExit(f"run slot {index} output section marker mismatch")
        end = offset + length
        if any(offset < other_end and other_start < end for other_start, other_end in sections):
            raise SystemExit(f"run slot {index} output section overlaps another receipt")
        sections.append((offset, end))
        if lane.startswith("default_"):
            default_ranges.append((offset, end))
    if len(set(default_ranges)) != 3:
        raise SystemExit("three distinct nonoverlapping default-parallel output sections are required")

    plan08_script = phase_dir / "verify-phase18-validation-evidence.sh"
    subprocess.run(("bash", str(plan08_script), "--final"), cwd=repo_root, check=True)
    prior = phase08_values(validation)
    if require(prior, "ci_disposition") != "not_added" or require(prior, "ci_command") != "none":
        raise SystemExit("Plan 08 CI disposition must remain ci_disposition=not_added and ci_command=none")
    current_ci = output(("git", "hash-object", ".github/workflows/ci.yml"))
    if require(prior, "ci_before_blob") != current_ci:
        raise SystemExit("Plan 08 CI blob changed from its recorded ci_before_blob")
    print("Phase 18 post-fix receipts and Plan 08 CI disposition verified")
PY
